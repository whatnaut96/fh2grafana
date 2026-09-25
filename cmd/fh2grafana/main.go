package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"os"
)

// Forza Horizon 6 has a packet size of 324 bytes. If FM 2023 comes in this will need updating.
const ExpectedPacketSize = 324

type forzaHorizon6Packet struct {
	// IsRaceOn is 1 when race/driving data is active and 0 in menus or stopped states.
	IsRaceOn int32

	// TimestampMS can eventually overflow back to 0.
	TimestampMS uint32

	// Engine RPM values.
	EngineMaxRPM     float32
	EngineIdleRPM    float32
	CurrentEngineRPM float32

	// Acceleration in the car's local space; X = right, Y = up, Z = forward.
	AccelerationX float32
	AccelerationY float32
	AccelerationZ float32

	// Velocity in the car's local space; X = right, Y = up, Z = forward.
	VelocityX float32
	VelocityY float32
	VelocityZ float32

	// Angular velocity in the car's local space, in rad/s; X = pitch, Y = yaw, Z = roll.
	AngularVelocityX float32
	AngularVelocityY float32
	AngularVelocityZ float32

	// Car orientation in radians.
	Yaw   float32
	Pitch float32
	Roll  float32

	// Suspension travel normalized: 0.0 = max stretch, 1.0 = max compression.
	NormalizedSuspensionTravelFrontLeft  float32
	NormalizedSuspensionTravelFrontRight float32
	NormalizedSuspensionTravelRearLeft   float32
	NormalizedSuspensionTravelRearRight  float32

	// Tire normalized slip ratio: 0 means 100% grip, |ratio| > 1.0 means loss of grip.
	TireSlipRatioFrontLeft  float32
	TireSlipRatioFrontRight float32
	TireSlipRatioRearLeft   float32
	TireSlipRatioRearRight  float32

	// Wheel rotation speed in rad/s.
	WheelRotationSpeedFrontLeft  float32
	WheelRotationSpeedFrontRight float32
	WheelRotationSpeedRearLeft   float32
	WheelRotationSpeedRearRight  float32

	// WheelOnRumbleStrip is 1 when the wheel is on a rumble strip and 0 otherwise.
	WheelOnRumbleStripFrontLeft  int32
	WheelOnRumbleStripFrontRight int32
	WheelOnRumbleStripRearLeft   int32
	WheelOnRumbleStripRearRight  int32

	// WheelInPuddle is 1 when the wheel is in a puddle and 0 otherwise.
	WheelInPuddleFrontLeft  int32
	WheelInPuddleFrontRight int32
	WheelInPuddleRearLeft   int32
	WheelInPuddleRearRight  int32

	// Non-dimensional surface rumble values passed to controller force feedback.
	SurfaceRumbleFrontLeft  float32
	SurfaceRumbleFrontRight float32
	SurfaceRumbleRearLeft   float32
	SurfaceRumbleRearRight  float32

	// Tire normalized slip angle: 0 means 100% grip, |angle| > 1.0 means loss of grip.
	TireSlipAngleFrontLeft  float32
	TireSlipAngleFrontRight float32
	TireSlipAngleRearLeft   float32
	TireSlipAngleRearRight  float32

	// Tire normalized combined slip: 0 means 100% grip, |slip| > 1.0 means loss of grip.
	TireCombinedSlipFrontLeft  float32
	TireCombinedSlipFrontRight float32
	TireCombinedSlipRearLeft   float32
	TireCombinedSlipRearRight  float32

	// Actual suspension travel in meters.
	SuspensionTravelMetersFrontLeft  float32
	SuspensionTravelMetersFrontRight float32
	SuspensionTravelMetersRearLeft   float32
	SuspensionTravelMetersRearRight  float32

	// Car identifiers and classification.
	CarOrdinal          int32
	CarClass            int32
	CarPerformanceIndex int32
	DriveTrainType      int32
	NumCylinders        int32
	CarGroup            uint32

	// Smashable object collision data.
	SmashableVelDiff float32
	SmashableMass    float32

	// Position in world space, in meters.
	PositionX float32
	PositionY float32
	PositionZ float32

	// Speed is meters per second, power is watts, torque is newton-meters.
	Speed  float32
	Power  float32
	Torque float32

	// Tire temperature.
	TireTempFrontLeft  float32
	TireTempFrontRight float32
	TireTempRearLeft   float32
	TireTempRearRight  float32

	// Boost is PSI above atmospheric. Fuel is 0.0 empty to 1.0 full.
	Boost float32
	Fuel  float32

	// Distance is meters. Lap and race times are seconds and 0.0 when not applicable.
	DistanceTraveled float32
	BestLap          float32
	LastLap          float32
	CurrentLap       float32
	CurrentRaceTime  float32

	// Race and player input state.
	LapNumber    uint16
	RacePosition uint8
	Acceleration uint8
	Brake        uint8
	Clutch       uint8
	Handbrake    uint8
	Gear         uint8

	// Signed inputs range from -127 to 127.
	Steer                       int8
	NormalizedDrivingLine       int8
	NormalizedAIBrakeDifference int8

	// The documented fields account for 323 bytes; FH6 packets include this trailing byte.
	Padding uint8
}

func parsePacket(rawBuffer []byte) (forzaHorizon6Packet, error) {
	var packet forzaHorizon6Packet
	if len(rawBuffer) != ExpectedPacketSize {
		return packet, fmt.Errorf("invalid packet size: got %d bytes, expected %d", len(rawBuffer), ExpectedPacketSize)
	}

	err := binary.Read(bytes.NewReader(rawBuffer), binary.LittleEndian, &packet)
	if err != nil {
		return packet, fmt.Errorf("parse FH6 packet: %w", err)
	}

	return packet, nil
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
		packet, err := parsePacket(buffer[:n])
		if err != nil {
			slog.Error("could not parse packet", "error", err)
			continue
		}
		slog.Info("testing IRO", "val", packet.IsRaceOn)
	}
}
