package coresystems

import (
	"math"

	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/tteokbokki/motion"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/TheBitDrifter/netcode_example/shared/actions"
	"github.com/TheBitDrifter/netcode_example/shared/components"
)

const (
	SPEED_X            = 120.0 // Player's horizontal movement speed
	SNAP_FORCE         = 40.0  // Downward force to keep player attached to slopes
	JUMP_FORCE         = 320.0 // Upward force applied when jumping
	COYOTE_TIME        = 10    // Ticks after leaving ground where jump is still allowed
	INPUT_BUFFER_TICKS = 5     // Ticks before landing where a jump input is remembered
)

// PlayerMovementSystem handles all player movement mechanics including horizontal
// movement on flat ground and slopes, jumping with coyote time + early jump buffering,
// and platform drop-through functionality.
type PlayerMovementSystem struct{}

func (sys PlayerMovementSystem) Run(scene blueprint.Scene, dt float64) error {
	sys.handleHorizontal(scene)
	sys.handleJump(scene)
	return sys.handleDown(scene)
}

// handleHorizontal processes left/right movement with different behaviors for:
// - Air movement
// - Flat ground movement
// - Uphill/downhill slope movement with proper tangent calculations
func (PlayerMovementSystem) handleHorizontal(scene blueprint.Scene) {
	cursor := scene.NewCursor(blueprint.Queries.ActionBuffer)
	currentTick := scene.CurrentTick()

	for range cursor.Next() {
		dyn := motion.Components.Dynamics.GetFromCursor(cursor)
		incomingActions := input.Components.ActionBuffer.GetFromCursor(cursor)
		direction := spatial.Components.Direction.GetFromCursor(cursor)

		_, pressedLeft := incomingActions.ConsumeAction(actions.Left)
		if pressedLeft {
			direction.SetLeft()
		}

		_, pressedRight := incomingActions.ConsumeAction(actions.Right)
		if pressedRight {
			direction.SetRight()
		}

		isMovingHorizontal := pressedLeft || pressedRight

		// Check ground status
		//
		// First check if the OnGroundComponent exists and get its value safely
		isGroundComponentPresent, onGround := components.OnGroundComponent.GetFromCursorSafe(cursor)
		isGrounded := isGroundComponentPresent && currentTick-1 == onGround.LastTouch

		// Default to airborne movement if no ground component exists
		if !isGrounded {
			if isMovingHorizontal {
				dyn.Vel.X = SPEED_X * direction.AsFloat()
			}
			continue
		}

		// Apply small downward force to keep player attached to slopes when grounded
		// Only applies if player has been on ground for a while and touched ground last tick
		dyn.Vel.Y = math.Max(dyn.Vel.Y, SNAP_FORCE) // Apply downward force

		// Check if player is on a flat surface (normal pointing straight up)
		flat := onGround.SlopeNormal.X == 0 && onGround.SlopeNormal.Y == 1
		if flat {
			// Same as air movement on flat ground
			if isMovingHorizontal {
				dyn.Vel.X = SPEED_X * direction.AsFloat()
			}
			// Skip slope handling
			continue
		}

		// Not flat or airborne (slope)
		if isMovingHorizontal {
			// Calculate tangent vector along the slope
			// The tangent is perpendicular to the normal, so we swap X/Y and negate Y
			tangent := vector.Two{X: onGround.SlopeNormal.Y, Y: -onGround.SlopeNormal.X}

			// Determine if player is moving uphill by checking if direction and normal X have same sign
			isUphill := (direction.AsFloat() * onGround.SlopeNormal.X) > 0

			// Scale tangent by movement direction for correct slope alignment
			slopeDir := tangent.Scale(direction.AsFloat())

			if isUphill {
				// When going uphill, only set X velocity and let physics handle Y
				dyn.Vel.X = slopeDir.X * SPEED_X
			} else {
				// When going downhill, help player follow the slope with both X and Y velocities
				dyn.Vel.X = slopeDir.X * SPEED_X
				dyn.Vel.Y = slopeDir.Y * SPEED_X
			}
		}
	}
}

// handleJump processes jump inputs with coyote time and input buffering features
// Coyote time: Player can jump shortly after leaving a platform
// Input buffering: Jump inputs are remembered and applied when landing
func (PlayerMovementSystem) handleJump(scene blueprint.Scene) {
	// Create query for players eligible to jump (have ground and input components)
	playersEligibleToJumpQuery := warehouse.Factory.NewQuery()
	playersEligibleToJumpQuery.And(components.OnGroundComponent, input.Components.ActionBuffer)

	// Get all entities that match the query
	cursor := scene.NewCursor(playersEligibleToJumpQuery)
	currentTick := scene.CurrentTick()

	for range cursor.Next() {
		// Get required components
		dyn := motion.Components.Dynamics.GetFromCursor(cursor)
		incomingActions := input.Components.ActionBuffer.GetFromCursor(cursor)
		jumpState := components.JumpStateComponent.GetFromCursor(cursor)

		// OnGroundComponent is guaranteed to exist because of our query
		onGround := components.OnGroundComponent.GetFromCursor(cursor)

		// Check for jump action
		if stampedAction, actionReceived := incomingActions.ConsumeAction(actions.Jump); actionReceived {

			// Coyote time: Allow jumping within certain ticks of leaving ground
			playerGroundedWithinCoyoteTime := currentTick-onGround.LastTouch <= COYOTE_TIME

			// Action buffering checks:
			//
			// 1. Was action received before touching ground?
			jumpIsBeforeGroundTouch := stampedAction.Tick <= onGround.LastTouch
			// 2. Was action within the buffer window?
			jumpWithinBufferWindow := onGround.LastTouch-stampedAction.Tick <= INPUT_BUFFER_TICKS
			// Combined buffer condition
			validBufferedJumpInput := jumpIsBeforeGroundTouch && jumpWithinBufferWindow

			// Direct jump: action received while already on ground
			directJumpAction := stampedAction.Tick >= onGround.LastTouch

			// Prevent double jumps: Make sure player hasn't jumped since last touching ground
			playerHasNotJumpedSinceGroundTouch := jumpState.LastJump < onGround.LastTouch

			// Player can jump if:
			// 1. They haven't jumped since touching ground, AND
			// 2a. They are in coyote time with a direct action, OR
			// 2b. They have a valid buffered action from before landing
			canJump := playerHasNotJumpedSinceGroundTouch &&
				((playerGroundedWithinCoyoteTime && directJumpAction) || validBufferedJumpInput)

			if canJump {
				// Apply upward velocity and acceleration for jump
				dyn.Vel.Y = -JUMP_FORCE
				dyn.Accel.Y = -JUMP_FORCE
				// Record jump time
				jumpState.LastJump = currentTick
			}
		}
	}
}

// handleDown processes down input for platform drop-through functionality
// This allows players to press down to fall through one-way platforms
func (PlayerMovementSystem) handleDown(scene blueprint.Scene) error {
	// Create query for players eligible to drop (have ground and input components)
	playersEligibleToDropQuery := warehouse.Factory.NewQuery()
	playersEligibleToDropQuery.And(components.OnGroundComponent, input.Components.ActionBuffer)

	cursor := scene.NewCursor(playersEligibleToDropQuery)
	currentTick := scene.CurrentTick()

	for range cursor.Next() {
		// OnGroundComponent is guaranteed to exist because of our query
		onGround := components.OnGroundComponent.GetFromCursor(cursor)
		jumpState := components.JumpStateComponent.GetFromCursor(cursor)

		// The component exists but were not actually grounded (will be removed soon)
		// Likely being held onto for coyote time tracking
		if onGround.LastTouch != currentTick-1 {
			continue
		}

		// Don't allow platform drop if player just jumped this tick
		if jumpState.LastJump == currentTick {
			continue
		}

		// Get player entity
		playerEntity, err := cursor.CurrentEntity()
		if err != nil {
			return err
		}

		incomingActions := input.Components.ActionBuffer.GetFromCursor(cursor)

		// Check for down action
		if stampedAction, inputReceived := incomingActions.ConsumeAction(actions.Down); inputReceived {
			// Only process inputs from this tick (ignore buffered inputs)
			if currentTick-stampedAction.Tick > 5 {
				continue
			}

			// Otherwise, add the IgnorePlatform component to allow dropping through platforms
			err := playerEntity.EnqueueAddComponent(components.IgnorePlatformComponent)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
