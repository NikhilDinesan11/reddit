// internal/actors/message_manager.go
package actors

import (
    "fmt"
    "github.com/asynkron/protoactor-go/actor"
    "redditclone/internal/messages"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func (state *MessageManagerActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *messages.SendDirectMessageMsg:
        messageID := fmt.Sprintf("msg_%d", len(state.Messages))
        newMessage := &messages.DirectMessage{
            Id:         messageID,
            FromUserId: msg.FromUserId,
            ToUserId:   msg.ToUserId,
            Content:    msg.Content,
            Timestamp:  timestamppb.Now(),
        }

        // Add message to recipient's inbox
        if _, exists := state.Messages[msg.ToUserId]; !exists {
            state.Messages[msg.ToUserId] = make([]*messages.DirectMessage, 0)
        }
        state.Messages[msg.ToUserId] = append(state.Messages[msg.ToUserId], newMessage)

        context.Respond(&messages.OperationResponse{
            Success: true,
            Id:      messageID,
            Result: &messages.OperationResponse_Message{
                Message: newMessage,
            },
        })
    }
}