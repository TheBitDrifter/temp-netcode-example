module github.com/TheBitDrifter/netcode_example/server

go 1.24.1

replace github.com/TheBitDrifter/bappa/table => ../../Bappa/table/

replace github.com/TheBitDrifter/bappa/warehouse => ../../Bappa/warehouse/

replace github.com/TheBitDrifter/bappa/tteokbokki => ../../Bappa/tteokbokki/

replace github.com/TheBitDrifter/bappa/blueprint => ../../Bappa/blueprint/

replace github.com/TheBitDrifter/bappa/coldbrew => ../../Bappa/coldbrew/

replace github.com/TheBitDrifter/bappa/environment => ../../Bappa/environment/

replace github.com/TheBitDrifter/bappa/drip => ../../Bappa/drip/

replace github.com/TheBitDrifter/netcode_example/shared => ../shared/

replace github.com/TheBitDrifter/netcode_example/sharedclient => ../sharedclient/

require (
	github.com/TheBitDrifter/bappa/blueprint v0.0.0-20250408214137-aae872bb6dfc
	github.com/TheBitDrifter/bappa/drip v0.0.0-00010101000000-000000000000
	github.com/TheBitDrifter/bappa/warehouse v0.0.0-20250408214137-aae872bb6dfc
	github.com/TheBitDrifter/netcode_example/shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/TheBitDrifter/bappa/environment v0.0.0-00010101000000-000000000000 // indirect
	github.com/TheBitDrifter/bappa/table v0.0.0-20250408214137-aae872bb6dfc // indirect
	github.com/TheBitDrifter/bappa/tteokbokki v0.0.0-20250408214137-aae872bb6dfc // indirect
	github.com/TheBitDrifter/bark v0.0.0-20250302175939-26104a815ed9 // indirect
	github.com/TheBitDrifter/mask v0.0.1-early-alpha.1 // indirect
	github.com/TheBitDrifter/util v0.0.0-20241102212109-342f4c0a810e // indirect
)
