// internal/actors/user_manager.go
package actors

import (
    "fmt"
    "github.com/asynkron/protoactor-go/actor"
    "redditclone/internal/messages"
)

func (state *UserManagerActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *actor.Started:
        if state.Users == nil {
            state.Users = make(map[string]*messages.User)
        }

    case *messages.RegisterUserMsg:
        // Check if username exists
        for _, user := range state.Users {
            if user.Username == msg.Username {
                if context.Sender() != nil {
                    context.Send(context.Sender(), &messages.OperationResponse{
                        Success: false,
                        Error: "username already exists",
                    })
                }
                return
            }
        }

        userID := fmt.Sprintf("user_%d", len(state.Users)+1)
        newUser := &messages.User{
            Id:       userID,
            Username: msg.Username,
            Karma:    0,
        }
        state.Users[userID] = newUser

        if context.Sender() != nil {
            context.Send(context.Sender(), &messages.OperationResponse{
                Success: true,
                Id:      userID,
                Result: &messages.OperationResponse_User{
                    User: newUser,
                },
            })
        }
    }
}