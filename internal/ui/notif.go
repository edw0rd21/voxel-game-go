package ui

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

type Notification struct {
	content     string
	createdAt   time.Time
	duration    time.Duration
	textElement *Text
}

type NotificationSystem struct {
	font          *Font
	notifications []Notification

	// Settings
	xPos, yPos   float32
	stackUpwards bool

	screenWidth  int
	screenHeight int
	baseScale    float32
}

func NewNotificationSystem(font *Font, screenWidth, screenHeight int) *NotificationSystem {
	return &NotificationSystem{
		font:          font,
		notifications: make([]Notification, 0),
		xPos:          20,
		yPos:          100,
		stackUpwards:  true,
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		baseScale:     1.0,
	}
}

func (ns *NotificationSystem) Init() error {
	return nil
}

func (ns *NotificationSystem) Update(state interface{}) {
	if screenSize, ok := state.(*ScreenSize); ok {
		if screenSize.Width != ns.screenWidth || screenSize.Height != ns.screenHeight {
			ns.screenWidth = screenSize.Width
			ns.screenHeight = screenSize.Height

			// Recalculate scale based on screen size
			scaleFactor := float32(screenSize.Width) / 1920.0 // Reference: 1920x1080
			newScale := ns.baseScale * scaleFactor

			// Update all existing notifications
			for i := range ns.notifications {
				ns.notifications[i].textElement.scale = newScale
				ns.notifications[i].textElement.needsUpdate = true
			}
		}
	}
	now := time.Now()
	active := make([]Notification, 0)

	for i := range ns.notifications {
		if now.Sub(ns.notifications[i].createdAt) < ns.notifications[i].duration {
			active = append(active, ns.notifications[i])
		} else {
			if ns.notifications[i].textElement != nil {
				ns.notifications[i].textElement.Cleanup()
			}
		}
	}

	const maxMessages = 5
	if len(active) > maxMessages {
		removeCount := len(active) - maxMessages

		for i := 0; i < removeCount; i++ {
			if active[i].textElement != nil {
				active[i].textElement.Cleanup()
			}
		}

		active = active[removeCount:]
	}

	currentY := ns.yPos
	for i := len(active) - 1; i >= 0; i-- {
		n := &active[i]
		// Update position for stacking effect
		if n.textElement.y != currentY {
			n.textElement.y = currentY
			n.textElement.needsUpdate = true
			n.textElement.Update(nil)
		}

		textHeight := ns.font.LineHeight * n.textElement.scale
		padding := textHeight * 0.2
		spacing := textHeight + padding

		// Move cursor for next message
		if ns.stackUpwards {
			currentY += spacing
		} else {
			currentY -= spacing
		}
	}
	ns.notifications = active
}

func (ns *NotificationSystem) Draw(shaderProgram uint32, projection mgl32.Mat4) {
	for _, n := range ns.notifications {
		n.textElement.Draw(shaderProgram, projection)
	}
}

func (ns *NotificationSystem) Cleanup() {
	for i := range ns.notifications {
		if ns.notifications[i].textElement != nil {
			ns.notifications[i].textElement.Cleanup()
		}
	}
	ns.notifications = nil
}

func (ns *NotificationSystem) Add(message string) {
	// Calculate current scale
	scaleFactor := float32(ns.screenWidth) / 1920.0
	currentScale := ns.baseScale * scaleFactor
	txt := NewText(ns.font, message, ns.xPos, ns.yPos, currentScale, mgl32.Vec3{1, 1, 1})
	txt.Init() // Create VBOs

	ns.notifications = append(ns.notifications, Notification{
		content:     message,
		createdAt:   time.Now(),
		duration:    3 * time.Second,
		textElement: txt,
	})
}
