package iface

import "github.com/mcp4go/mcp4go/protocol"

type EventBus struct {
	PromptListChangedNotificationChan chan protocol.PromptListChangedNotification

	ResourceUpdatedNotificationChan     chan protocol.ResourceUpdatedNotification
	ResourceListChangedNotificationChan chan protocol.ResourceListChangedNotification

	ToolListChangedNotificationChan chan protocol.ToolListChangedNotification

	LoggingMessageNotificationChan chan protocol.LoggingMessageNotification
}

func NewEventBus() EventBus {
	return EventBus{
		PromptListChangedNotificationChan:   make(chan protocol.PromptListChangedNotification, 32),
		ResourceUpdatedNotificationChan:     make(chan protocol.ResourceUpdatedNotification, 32),
		ResourceListChangedNotificationChan: make(chan protocol.ResourceListChangedNotification, 32),
		ToolListChangedNotificationChan:     make(chan protocol.ToolListChangedNotification, 32),
		LoggingMessageNotificationChan:      make(chan protocol.LoggingMessageNotification, 32),
	}
}
