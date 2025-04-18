package coresystems

import (
	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/TheBitDrifter/netcode_example/shared/components"
)

type OnGroundClearingSystem struct{}

func (OnGroundClearingSystem) Run(scene blueprint.Scene, dt float64) error {
	// 15 is a bit of a magic number— what matters is that the value is greater than coyote timer
	// Systems should also check lastTouch alongside presence to avoid odd behavior
	const expirationTicks = 15

	// Query any entity that has onGround
	onGroundQuery := warehouse.Factory.NewQuery().And(components.OnGroundComponent)
	onGroundCursor := scene.NewCursor(onGroundQuery)

	// Iterate through matched entities
	for range onGroundCursor.Next() {
		// Get the onGround component state
		onGround := components.OnGroundComponent.GetFromCursor(onGroundCursor)

		// If it's expired, remove it
		if scene.CurrentTick()-onGround.LastTouch > expirationTicks {
			groundedEntity, _ := onGroundCursor.CurrentEntity()

			// We can't mutate while iterating so we enqueue the changes instead
			err := groundedEntity.EnqueueRemoveComponent(components.OnGroundComponent)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
