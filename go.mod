module github.com/hashicorp/go-msgpack

go 1.19

// v1.1.5 merged upstream ugorji/go which breaks compatibility with previous versions.
// v2 was created to safely merge upstream changes but allow projects relying on
// older go-msgpack modules to maintain their pinned versions.
retract (
	v1.1.6 // Contains v1.1.5 retraction only.
	v1.1.5
)
