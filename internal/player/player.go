package player

import (
	"math"

	"voxel-game/internal/camera"
	"voxel-game/internal/world"

	"github.com/go-gl/mathgl/mgl32"
)

type TargetBlock struct {
	Hit  bool
	Pos  mgl32.Vec3
	Face int
}

type Player struct {
	camera *camera.Camera
	world  *world.World

	PhysicsPos mgl32.Vec3

	walkSpeed float32
	jumpForce float32
	velocity  mgl32.Vec3

	grounded bool
	width    float32
	height   float32

	target TargetBlock

	walkingTime float32
}

func NewPlayer(cam *camera.Camera, w *world.World) *Player {
	p := &Player{
		camera:     cam,
		world:      w,
		PhysicsPos: cam.Position,
		width:      0.6,
		height:     1.8,
		walkSpeed:  4.3,
		jumpForce:  8.0,
	}
	p.camera.Position = p.PhysicsPos.Add(mgl32.Vec3{0, p.GetEyeHeight(), 0})
	return p
}

func (p *Player) Update(deltaTime float32) {
	const gravity = 25.0
	const terminalVelocity = -50.0

	// Apply velocity
	movement := p.velocity.Mul(deltaTime)
	newPos := p.PhysicsPos.Add(movement)

	// Collision detection
	finalPos := p.handleCollision(newPos, &p.velocity)
	p.PhysicsPos = finalPos

	// Check if grounded
	p.grounded = p.isGrounded()

	// Apply gravity
	if !p.grounded {
		p.velocity[1] -= gravity * deltaTime
		if p.velocity[1] < terminalVelocity {
			p.velocity[1] = terminalVelocity
		}
	} else {
		// velocity is zero when grounded to prevent accumulation
		if p.velocity[1] < 0 {
			p.velocity[1] = 0
		}
	}

	// Damping
	friction := float32(10.0)
	if !p.grounded {
		friction = 1.0 // Low friction in air (air control)
	}

	dragFactor := float32(1.0) - (friction * deltaTime)
	if dragFactor < 0 {
		dragFactor = 0
	}

	p.velocity[0] *= dragFactor
	p.velocity[2] *= dragFactor

	if mgl32.Abs(p.velocity[0]) < 0.1 {
		p.velocity[0] = 0
	}
	if mgl32.Abs(p.velocity[2]) < 0.1 {
		p.velocity[2] = 0
	}

	// View bobbing
	horizontalSpeed := float32(math.Sqrt(float64(p.velocity[0]*p.velocity[0] + p.velocity[2]*p.velocity[2])))

	if p.grounded && horizontalSpeed > 0.1 {
		p.walkingTime += deltaTime * 10.0
	} else {
		p.walkingTime = 0
	}

	bobOffsetY := float32(math.Sin(float64(p.walkingTime))) * 0.1
	bobOffsetX := float32(math.Sin(float64(p.walkingTime/2.0))) * 0.05

	p.camera.Position = p.PhysicsPos.Add(mgl32.Vec3{0, p.GetEyeHeight(), 0})

	p.camera.Position[1] += bobOffsetY
	sway := p.camera.Right.Mul(bobOffsetX)
	p.camera.Position = p.camera.Position.Add(sway)

	p.updateTarget()
}

func (p *Player) updateTarget() {
	hit, x, y, z, face := p.Raycast(5.0)
	if hit {
		p.target = TargetBlock{
			Hit:  true,
			Pos:  mgl32.Vec3{float32(x), float32(y), float32(z)},
			Face: face,
		}
	} else {
		p.target = TargetBlock{}
	}
}

func (p *Player) TargetBlock() TargetBlock {
	return p.target
}

func (p *Player) Move(direction mgl32.Vec3, deltaTime float32) {
	if direction.Len() > 0 {
		accel := float32(60.0)
		if !p.grounded {
			accel = 10.0 // Slower acceleration in air
		}

		p.velocity = p.velocity.Add(direction.Mul(accel * deltaTime))

		flatVel := mgl32.Vec3{p.velocity[0], 0, p.velocity[2]}
		if flatVel.Len() > p.walkSpeed {
			flatVel = flatVel.Normalize().Mul(p.walkSpeed)
			p.velocity[0] = flatVel[0]
			p.velocity[2] = flatVel[2]
		}
	}
}

func (p *Player) Jump() {
	if p.grounded {
		p.velocity[1] = p.jumpForce
		p.grounded = false // Instant feedback
	}
}

func (p *Player) handleCollision(newPos mgl32.Vec3, velocity *mgl32.Vec3) mgl32.Vec3 {
	// Simple AABB collision
	testPos := mgl32.Vec3{newPos[0], p.PhysicsPos[1], p.PhysicsPos[2]}
	if p.checkCollision(testPos) {
		newPos[0] = p.PhysicsPos[0] // Revert X
		velocity[0] = 0             // Stop X momentum
	}

	testPos = mgl32.Vec3{newPos[0], p.PhysicsPos[1], newPos[2]}
	if p.checkCollision(testPos) {
		newPos[2] = p.PhysicsPos[2] // Revert Z
		velocity[2] = 0             // Stop Z momentum
	}

	testPos = mgl32.Vec3{newPos[0], newPos[1], newPos[2]}
	if p.checkCollision(testPos) {
		newPos[1] = p.PhysicsPos[1]

		if velocity[1] < 0 {
			p.grounded = true
		}

		velocity[1] = 0 // Stop vertical momentum
	}

	return newPos
}

func (p *Player) checkCollision(pos mgl32.Vec3) bool {
	minX := int(math.Floor(float64(pos[0] - p.width/2)))
	maxX := int(math.Floor(float64(pos[0] + p.width/2)))
	minY := int(math.Floor(float64(pos[1])))
	maxY := int(math.Floor(float64(pos[1] + p.height)))
	minZ := int(math.Floor(float64(pos[2] - p.width/2)))
	maxZ := int(math.Floor(float64(pos[2] + p.width/2)))

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				if p.world.GetBlock(x, y, z) != world.BlockAir {
					return true
				}
			}
		}
	}
	return false
}

func (p *Player) isGrounded() bool {
	minX := int(math.Floor(float64(p.PhysicsPos[0] - p.width/2)))
	maxX := int(math.Floor(float64(p.PhysicsPos[0] + p.width/2)))
	minZ := int(math.Floor(float64(p.PhysicsPos[2] - p.width/2)))
	maxZ := int(math.Floor(float64(p.PhysicsPos[2] + p.width/2)))

	checkY := int(math.Floor(float64(p.PhysicsPos[1]))) - 1
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			if p.world.GetBlock(x, checkY, z) != world.BlockAir {
				return true
			}
		}
	}

	return false
}

// Raycast to find the block the player is looking at
func (p *Player) Raycast(maxDistance float32) (hit bool, x, y, z int, face int) {
	pos := p.camera.Position
	dir := p.camera.Front
	step := float32(0.1)

	for dist := float32(0); dist < maxDistance; dist += step {
		checkPos := pos.Add(dir.Mul(dist))
		bx := int(math.Floor(float64(checkPos[0])))
		by := int(math.Floor(float64(checkPos[1])))
		bz := int(math.Floor(float64(checkPos[2])))

		if p.world.GetBlock(bx, by, bz) != world.BlockAir {
			// Determine which face was hit
			prevPos := pos.Add(dir.Mul(dist - step))
			px := int(math.Floor(float64(prevPos[0])))
			py := int(math.Floor(float64(prevPos[1])))
			pz := int(math.Floor(float64(prevPos[2])))

			if bx != px {
				if bx > px {
					face = 3 // -X
				} else {
					face = 2 // +X
				}
			} else if by != py {
				if by > py {
					face = 5 // -Y
				} else {
					face = 4 // +Y
				}
			} else if bz != pz {
				if bz > pz {
					face = 1 // -Z
				} else {
					face = 0 // +Z
				}
			}

			return true, bx, by, bz, face
		}
	}

	return false, 0, 0, 0, 0
}

func (p *Player) BreakBlock() {
	if !p.target.Hit {
		return
	}

	pos := p.target.Pos
	p.world.SetBlock(
		int(pos.X()),
		int(pos.Y()),
		int(pos.Z()),
		world.BlockAir,
	)
}

func (p *Player) PlaceBlock(blockType world.BlockType) {
	if !p.target.Hit {
		return
	}

	x := int(p.target.Pos.X())
	y := int(p.target.Pos.Y())
	z := int(p.target.Pos.Z())

	switch p.target.Face {
	case 0:
		z++
	case 1:
		z--
	case 2:
		x++
	case 3:
		x--
	case 4:
		y++
	case 5:
		y--
	}

	if p.collidesWithPlayer(float32(x), float32(y), float32(z)) {
		return
	}

	p.world.SetBlock(x, y, z, blockType)

}

func (p *Player) collidesWithPlayer(x, y, z float32) bool {
	px := p.PhysicsPos.X()
	py := p.PhysicsPos.Y()
	pz := p.PhysicsPos.Z()

	return mgl32.Abs(px-x) < p.width &&
		py < y+p.height &&
		py+p.height > y &&
		mgl32.Abs(pz-z) < p.width
}
func (p *Player) GetEyeHeight() float32 {
	return p.height - 0.2
}

func (p *Player) TeleportToCamera() {
	eyeOffset := mgl32.Vec3{0, p.GetEyeHeight(), 0}

	p.PhysicsPos = p.camera.Position.Sub(eyeOffset)

	p.velocity = mgl32.Vec3{0, 0, 0}
	p.grounded = false
}
