package wheeliegoclient

type ConnectStatusCode int32

const (
	SuccessConnect ConnectStatusCode = iota
	ConnectServerError
	TooManyConnections
)

type AuthStatusCode int32

const (
	SuccessAuth AuthStatusCode = iota
	InvalidCredentials
	AuthParsingError
	AuthServerError
)

type GetStatusCode int32

const (
	SuccessGet GetStatusCode = iota
	GetParsingError
	GetClientProblem
	GetServerError
)

type CloseStatusCode int32

const (
	CloseClientAppShutdown CloseStatusCode = iota
	CloseClientInternalError
	CloseLogout
	CloseNoPingResponse
	CloseNotMaintainedProtocolVersion
	CloseServerPacketParsing
	CloseOtherError
)

type DisconnectStatusCode int32

const (
	DisconnectServerApplicationShutdown DisconnectStatusCode = iota
	DisconnectServerInternalError
	DisconnectNoPingPacket
	DisconnectNotMaintainedProtocolVersion
	DisconnectClientPacketParsingProblem
	DisconnectFailedConnectionSetup
	DisconnectAuthFailed
	DisconnectOtherError
)
