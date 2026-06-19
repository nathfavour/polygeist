module github.com/nathfavour/polygeist

go 1.25.5

require (
	github.com/gorilla/websocket v1.5.3
	github.com/nathfavour/anyisland v0.0.0
	github.com/nathfavour/auracrab v0.0.0
	github.com/nathfavour/vibeauracle/pkg/engine v0.0.0
)

require github.com/nathfavour/vibeauracle/pkg/ipc v0.0.0 // indirect

replace github.com/nathfavour/anyisland => ./anyisland

replace github.com/nathfavour/auracrab => ./auracrab

replace github.com/nathfavour/vibeauracle/pkg/engine => ./vibeauracle/pkg/engine

replace github.com/nathfavour/vibeauracle/pkg/ipc => ./vibeauracle/pkg/ipc
