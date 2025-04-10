package handlers

import (
	"context"
	"encoding/json"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// Handle logging/setLevel request
type SetLevelHandler struct {
	// Current log level
	currentLevel protocol.LoggingLevel
}

// Create a new instance
func NewSetLevelHandler() *SetLevelHandler {
	return &SetLevelHandler{
		currentLevel: protocol.LoggingLevelInfo, // Default log level is Info
	}
}

// Handle logging/setLevel request
func (x *SetLevelHandler) Handle(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.SetLevelRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Set new log level
	x.currentLevel = req.Level

	result := protocol.SetLevelResult{}
	return json.Marshal(result)
}

// Returns the result
func (x *SetLevelHandler) Method() protocol.McpMethod {
	return protocol.MethodSetLevel
}

// Get the specified data
func (x *SetLevelHandler) GetCurrentLevel() protocol.LoggingLevel {
	return x.currentLevel
}

// LoggingMessageSender used for sending log messages
type LoggingMessageSender struct {
	logChan      chan<- protocol.LoggingMessageNotification
	levelHandler *SetLevelHandler
}

// Create a new instance
func NewLoggingMessageSender(
	bus iface.EventBus,
	levelHandler *SetLevelHandler,
) *LoggingMessageSender {
	//nolint:whitespace
	return &LoggingMessageSender{
		logChan:      bus.LoggingMessageNotificationChan,
		levelHandler: levelHandler,
	}
}

// LogLevelMap maps log levels for comparison
var LogLevelMap = map[protocol.LoggingLevel]int{
	protocol.LoggingLevelDebug:     0,
	protocol.LoggingLevelInfo:      1,
	protocol.LoggingLevelNotice:    2,
	protocol.LoggingLevelWarning:   3,
	protocol.LoggingLevelError:     4,
	protocol.LoggingLevelCritical:  5,
	protocol.LoggingLevelAlert:     6,
	protocol.LoggingLevelEmergency: 7,
}

// SendLogMessage sends log messages, but won't send if the message level is below the current level setting
func (x *LoggingMessageSender) SendLogMessage(level protocol.LoggingLevel, logger string, data interface{}) error {
	// Check if the message level should be sent
	currentLevel := x.levelHandler.GetCurrentLevel()

	if LogLevelMap[level] < LogLevelMap[currentLevel] {
		// Message level is below the current level, don't send
		return nil
	}

	bs, _ := json.Marshal(data)
	x.logChan <- protocol.LoggingMessageNotification{
		Level:  level,
		Logger: logger,
		Data:   bs,
	}

	return nil
}

// Convenience methods for sending logs at different levels
func (x *LoggingMessageSender) Debug(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelDebug, logger, data)
}

func (x *LoggingMessageSender) Info(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelInfo, logger, data)
}

func (x *LoggingMessageSender) Notice(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelNotice, logger, data)
}

func (x *LoggingMessageSender) Warning(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelWarning, logger, data)
}

func (x *LoggingMessageSender) Error(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelError, logger, data)
}

func (x *LoggingMessageSender) Critical(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelCritical, logger, data)
}

func (x *LoggingMessageSender) Alert(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelAlert, logger, data)
}

func (x *LoggingMessageSender) Emergency(logger string, data interface{}) error {
	return x.SendLogMessage(protocol.LoggingLevelEmergency, logger, data)
}
