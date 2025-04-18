package clientsystems

import (
	"math"

	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/tteokbokki/motion"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/TheBitDrifter/netcode_example/shared/components"
	"github.com/TheBitDrifter/netcode_example/shared/sounds"
)

type PlayerSoundSystem struct{}

func (sys PlayerSoundSystem) Run(cli coldbrew.LocalClient, scene coldbrew.Scene) error {
	playersWithSoundsOnTheGround := warehouse.Factory.NewQuery().And(
		client.Components.SoundBundle,
		input.Components.ActionBuffer,
		motion.Components.Dynamics,
		components.OnGroundComponent,
	)

	cursor := scene.NewCursor(playersWithSoundsOnTheGround)

	for range cursor.Next() {

		// Get state
		soundBundle := client.Components.SoundBundle.GetFromCursor(cursor)
		dyn := motion.Components.Dynamics.GetFromCursor(cursor)
		onGround := components.OnGroundComponent.GetFromCursor(cursor)
		jumpState := components.JumpStateComponent.GetFromCursor(cursor)
		currentTick := scene.CurrentTick()

		// Landed sound
		if onGround.Landed == currentTick {
			landingSound, _ := coldbrew.MaterializeSound(soundBundle, sounds.Land)
			player := landingSound.GetAny()

			// A hack to prevent landing sound artifacts between scenes
			// In a more robust setup, we might track if a player has recently changed scenes via a component
			// Such a component would be helpful here
			sceneRecentlySelected := scene.CurrentTick()-scene.LastSelectedTick() < 30

			if !player.IsPlaying() && !sceneRecentlySelected {
				player.Rewind()
				player.Play()
			}
		}

		// Jump sound

		// Gotta have y velocity
		if dyn.Vel.Y < 5 && jumpState.LastJump == currentTick {
			jumpSound, _ := coldbrew.MaterializeSound(soundBundle, sounds.Jump)
			player := jumpSound.GetAny()

			if !player.IsPlaying() {
				player.Rewind()
				player.Play()
			}
		}

		// Run Sound
		// Must be moving horizontally
		const minMovementSpeed = 20.0
		if math.Abs(dyn.Vel.X) <= minMovementSpeed {
			continue
		}
		// Ensure onGround is not just available for coyote timer
		touchedGroundThisTick := onGround.LastTouch == currentTick
		if !touchedGroundThisTick {
			continue
		}

		runSound, err := coldbrew.MaterializeSound(soundBundle, sounds.Run)
		if err != nil {
			return err
		}
		player := runSound.GetAny()

		if !player.IsPlaying() {
			player.Rewind()
			player.Play()
		}
	}

	return nil
}
