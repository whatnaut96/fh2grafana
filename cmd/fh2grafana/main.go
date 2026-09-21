// TODO: Packet comments copy from actual page: https://support.forza.net/hc/en-us/articles/51744149102611-Forza-Horizon-6-Data-Out-Documentation

package main

import (
	"encoding/binary"
	"log/slog"
	"math"
	"net"
	"os"
)

// Forza Horizon 6 has a packet size of 324 bytes. If FM 2023 comes in this will need updating.
const ExpectedPacketSize = 324

type forzaHorizon6Packet struct {
	IsRaceOn                             int32
	TimestampMS                          uint32
	EngineMaxRPM                         float32
	EngineIdleRPM                        float32
	CurrentEngineRPM                     float32
	VelocityX                            float32
	VelocityY                            float32
	VelocityZ                            float32
	AngularVelocityX                     float32
	AngularVelocityY                     float32
	AngularVelocityZ                     float32
	Yaw                                  float32
	Pitch                                float32
	Roll                                 float32
	NormalizedSuspensionTravelFrontLeft  float32
	NormalizedSuspensionTravelFrontRight float32
	NormalizedSuspensionTravelRearLeft   float32
	NormalizedSuspensionTravelRearRight  float32
	TireSlipRatioFrontLeft               float32
	TireSlipRatioFrontRight              float32
	TireSlipratioRearLeft                float32
	TireSlipRatioRearRight               float32
	WheelOnRumbleStripFrontLeft          int32
	WheelOnRumbleStripFrontRight         int32
	WheelOnRumbleStripRearLeft           int32
	WheelOnRumbleStripRearRight          int32

	TireCombinedSlipFrontLeft        float32
	TireCombinedSlipFrontRight       float32
	TireCombinedSlipRearLeft         float32
	TireCombinedSlipRearRight        float32
	SuspensionTravelMetersFrontLeft  float32
	SuspensionTravelMetersFrontRight float32
	SuspensionTravelMetersRearLeft   float32
	SuspensionTravelMetersRearRight  float32
	CarOrdinal                       int32
	CarClass                         int32
	CarPerformanceIndex              int32
	DriveTrainType                   int32
	NumCylinders                     int32
	CarGroup                         uint32
	SmashableVelDiff                 float32
	SmashableMass                    float32
	PositionX                        float32
	PositionY                        float32
	PositionZ                        float32
	Speed                            float32
	Power                            float32
	Torque                           float32
	TireTempFrontLeft                float32
	TireTempFrontRight               float32
	TireTempRearLeft                 float32
	TireTempRearRight                float32
	Boost                            float32

	// This is unused in FH games. More important if FM 2023 becomes a target
	Fuel float32

	DistanceTraveled            float32
	BestLap                     float32
	LastLap                     float32
	CurrentLap                  float32
	CurrentRaceTime             float32
	LapNumber                   uint16
	Acceleration                uint8
	Brake                       uint8
	Clutch                      uint8
	Handbrake                   uint8
	Gear                        uint8
	Steer                       uint8
	NormalizedDrivingLine       int8
	NormalizedAIBrakeDifference int8
}

func parsePacket(rawBuffer []byte) (forzaHorizon6Packet, error) {
	isRaceOn := int32(binary.LittleEndian.Uint32(rawBuffer[0:4]))
	timestamp := binary.LittleEndian.Uint32(rawBuffer[4:8])
	engineMaxRPM := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[8:12]),
	)

	engineIdleRPM := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[12:16]),
	)

	currentEngineRPM := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[16:20]),
	)

	velocityX := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[20:24]),
	)

	velocityY := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[24:28]),
	)

	angularVelocityX := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[28:32]),
	)

	angularVelocityY := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[32:36]),
	)

	angularVelocityZ := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[36:40]),
	)

	yaw := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[40:44]),
	)

	pitch := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[44:48]),
	)

	roll := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[48:52]),
	)

	// not typing all this out for these next names.
	// see above for what they are and where the usage is in the packet
	nstfl := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[52:56]),
	)

	nstfr := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[56:60]),
	)

	nstrl := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[60:64]),
	)

	nstrr := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[64:68]),
	)

	tsrfl := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[68:72]),
	)

	tsrfr := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[72:76]),
	)

	tsrrl := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[76:80]),
	)

	tsrrr := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[80:84]),
	)

	rumbleStripFL := int32(binary.LittleEndian.Uint32(rawBuffer[84:88]))
	rumbleStripFR := int32(binary.LittleEndian.Uint32(rawBuffer[88:92]))
	rumbleStripRL := int32(binary.LittleEndian.Uint32(rawBuffer[92:96]))
	rumbleStripRR := int32(binary.LittleEndian.Uint32(rawBuffer[96:100]))

	// NOTE: The gap occurs here because I don't need the rumble fields.
	// Going back to short names after our glimpse of light above
	tcsfl := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[116:120]),
	)

	tcsfr := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[120:124]),
	)

	tcsrl := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[124:128]),
	)

	tcsrr := math.Float32frombits(
		binary.LittleEndian.Uint32(rawBuffer[128:132]),
	)

	returnPacket := forzaHorizon6Packet{
		IsRaceOn:                             isRaceOn,
		TimestampMS:                          timestamp,
		EngineMaxRPM:                         engineMaxRPM,
		EngineIdleRPM:                        engineIdleRPM,
		CurrentEngineRPM:                     currentEngineRPM,
		VelocityX:                            velocityX,
		VelocityY:                            velocityY,
		AngularVelocityX:                     angularVelocityX,
		AngularVelocityY:                     angularVelocityY,
		AngularVelocityZ:                     angularVelocityZ,
		Yaw:                                  yaw,
		Pitch:                                pitch,
		Roll:                                 roll,
		NormalizedSuspensionTravelFrontLeft:  nstfl,
		NormalizedSuspensionTravelFrontRight: nstfr,
		NormalizedSuspensionTravelRearLeft:   nstrl,
		NormalizedSuspensionTravelRearRight:  nstrr,
		TireSlipRatioFrontLeft:               tsrfl,
		TireSlipRatioFrontRight:              tsrfr,
		TireSlipratioRearLeft:                tsrrl,
		TireSlipRatioRearRight:               tsrrr,
		WheelOnRumbleStripFrontLeft:          rumbleStripFL,
		WheelOnRumbleStripFrontRight:         rumbleStripFR,
		WheelOnRumbleStripRearLeft:           rumbleStripRL,
		WheelOnRumbleStripRearRight:          rumbleStripRR,
		TireCombinedSlipFrontLeft:            tcsfl,
		TireCombinedSlipFrontRight:           tcsfr,
		TireCombinedSlipRearLeft:             tcsrl,
		TireCombinedSlipRearRight:            tcsrr,
	}

	return returnPacket, nil
}

func main() {
	slog.Info("neom neom")
	localAddr, err := net.ResolveUDPAddr("udp", ":5300")
	if err != nil {
		slog.Error("failed to resolve local udp address", "error", err)
		os.Exit(1)
	}

	udpConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		slog.Error("could not start listening for udp", "address", localAddr.String(), "error", err)
	}
	defer udpConn.Close()

	slog.Info("UDP Server listening on port 5300")

	buffer := make([]byte, 324)
	for {
		n, _, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			slog.Error("could not read from UDP", "error", err)
			continue
		}

		if n != ExpectedPacketSize {
			slog.Error("invalid packet size", "received", n, "expected", ExpectedPacketSize)
			continue
		}
		packet, err := parsePacket(buffer)
		slog.Info("testing IRO", "val", packet.IsRaceOn)
	}
}
