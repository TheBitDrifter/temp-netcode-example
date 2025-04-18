package coresystems

import (
	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/tteokbokki/motion"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/TheBitDrifter/netcode_example/shared/components"
)

type PlayerBlockCollisionSystem struct{}

func (s PlayerBlockCollisionSystem) Run(scene blueprint.Scene, dt float64) error {
	blockTerrainQuery := warehouse.Factory.NewQuery().And(components.BlockTerrainTag)
	blockTerrainCursor := scene.NewCursor(blockTerrainQuery)
	playerCursor := scene.NewCursor(blueprint.Queries.ActionBuffer)

	// Outer loop is blocks
	for range blockTerrainCursor.Next() {
		// Inner is players
		for range playerCursor.Next() {
			// Delegate to helper
			err := s.resolve(scene, blockTerrainCursor, playerCursor) // Now pass in the scene
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (PlayerBlockCollisionSystem) resolve(scene blueprint.Scene, blockCursor, playerCursor *warehouse.Cursor) error {
	// Get the player pos, shape, and dynamics
	playerPosition := spatial.Components.Position.GetFromCursor(playerCursor)
	playerShape := spatial.Components.Shape.GetFromCursor(playerCursor)
	playerDynamics := motion.Components.Dynamics.GetFromCursor(playerCursor)

	// Get the block pos, shape, and dynamics
	blockPosition := spatial.Components.Position.GetFromCursor(blockCursor)
	blockShape := spatial.Components.Shape.GetFromCursor(blockCursor)
	blockDynamics := motion.Components.Dynamics.GetFromCursor(blockCursor)

	// Check grounded state
	playerAlreadyGrounded, onGround := components.OnGroundComponent.GetFromCursorSafe(playerCursor)

	// Check for a collision
	if ok, collisionResult := spatial.Detector.Check(
		*playerShape, *blockShape, playerPosition.Two, blockPosition.Two,
	); ok {

		playerOnTopOfBlock := collisionResult.IsTopB()
		blockOnTopOfPlayer := collisionResult.IsTop()

		// Determine if ground is sloped
		n := collisionResult.Normal
		horizontal := n.X == 0 && n.Y == 1 || n.X == 0 && n.Y == -1
		vertical := n.X == -1 && n.Y == 0 || n.X == 1 && n.Y == 0
		isSloped := !horizontal && !vertical

		if playerAlreadyGrounded && onGround.LastTouch != scene.CurrentTick()-1 {
			playerAlreadyGrounded = false
		}

		// Prevents snapping on AAB corner transitions/collisions
		if playerOnTopOfBlock && playerDynamics.Vel.Y < 0 && !isSloped {
			return nil
		}
		if blockOnTopOfPlayer && playerDynamics.Vel.Y > 0 {
			return nil
		}

		// Vertical resolver to prevent positional sliding on slopes
		if isSloped {
			motion.VerticalResolver.Resolve(
				&playerPosition.Two,
				&blockPosition.Two,
				playerDynamics,
				blockDynamics,
				collisionResult,
			)
		} else {
			// Otherwise resolve as normal
			motion.Resolver.Resolve(
				&playerPosition.Two,
				&blockPosition.Two,
				playerDynamics,
				blockDynamics,
				collisionResult,
			)
		}
		currentTick := scene.CurrentTick()

		isWall := !playerOnTopOfBlock && !blockOnTopOfPlayer
		if isWall {
			wallDirectionV := spatial.NewDirectionLeft()
			wallDirection := &wallDirectionV

			if blockPosition.X > playerPosition.X {
				wallDirection.SetRight()
			}

		}

		// Ensure the player is on top of the terrain before marking them as grounded
		if !playerOnTopOfBlock {
			return nil
		}
		// Update onGround accordingly (create or update)
		if !playerAlreadyGrounded {
			playerEntity, err := playerCursor.CurrentEntity()
			if err != nil {
				return err
			}
			// We cannot mutate during a cursor iteration, so we use the enqueue API
			err = playerEntity.EnqueueAddComponentWithValue(
				components.OnGroundComponent,
				components.OnGround{LastTouch: currentTick, Landed: currentTick, SlopeNormal: collisionResult.Normal},
			)
			if err != nil {
				return err
			}
		} else {
			onGround.LastTouch = scene.CurrentTick()
			onGround.SlopeNormal = collisionResult.Normal
		}

	}
	return nil
}
