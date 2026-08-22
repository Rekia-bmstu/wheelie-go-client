package wheeliegoclient

type PacketTag int32

const (
	FireAndForget PacketTag = iota
	Connect
	ConnectResponse
	Setup
	SetupResponse
	Auth
	AuthResponse
	Ping
	PingResponse
	Close
	Disconnect
	Get
	GetResponse
	CriticalGet
	CriticalGetResponse
	ServerPush
	ServerPushAck
	CriticalServerPush
	CriticalServerPushAck
	MultiPart
	MultiPartAck
	MultiGet
	MultiGetResponse
	MultiPush
	MultiPushResponse
)

type Priority int32

const (
	LowPriority Priority = iota
	MediumPriority
	HighPriority
)

var packetsPriorities = map[PacketTag]Priority{}
