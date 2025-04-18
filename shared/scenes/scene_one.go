package scenes

import (
	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/TheBitDrifter/netcode_example/shared/ldtk"
	"github.com/TheBitDrifter/netcode_example/shared/sounds"
)

const SCENE_ONE_NAME = "Scene1"

var SCENE_ONE_PRELOADED_ASSETS = client.NewPreLoadBlueprint().
	AddSprite(PLAYER_SPRITE_SHEET_PATH).
	AddSound(sounds.Jump).
	AddSound(sounds.Land).
	AddSound(sounds.Run)

var SceneOne = Scene{
	Name:    SCENE_ONE_NAME,
	Plan:    sceneOnePlan,
	Width:   ldtk.DATA.WidthFor(SCENE_ONE_NAME),
	Height:  ldtk.DATA.HeightFor(SCENE_ONE_NAME),
	Preload: *SCENE_ONE_PRELOADED_ASSETS,
}

func sceneOnePlan(width, height int, sto warehouse.Storage) error {
	// Load the image tiles
	err := ldtk.DATA.LoadTiles(SCENE_ONE_NAME, sto)
	if err != nil {
		return err
	}

	// Load the terrain
	// Pass the terrain archetypes in order of int grid layer they map to
	blockArchetype, _ := sto.NewOrExistingArchetype(BlockTerrainComposition...)
	platArchetype, _ := sto.NewOrExistingArchetype(PlatformComposition...)
	transferArchetype, _ := sto.NewOrExistingArchetype(CollisionPlayerTransferComposition...)

	err = ldtk.DATA.LoadIntGrid(SCENE_ONE_NAME, sto, blockArchetype, platArchetype, transferArchetype)
	if err != nil {
		return err
	}

	// Load custom LDTK entities
	err = ldtk.DATA.LoadEntities(SCENE_ONE_NAME, sto, entityRegistry)
	if err != nil {
		return err
	}

	// Music
	err = NewJazzMusic(sto)
	if err != nil {
		return err
	}

	return NewCityBackground(sto)
}
