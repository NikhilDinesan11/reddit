// internal/actors/subreddit_manager.go
package actors

import (
    "fmt"
    "github.com/asynkron/protoactor-go/actor"
    "redditclone/internal/messages"
)

func (state *SubRedditManagerActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *messages.CreateSubRedditMsg:
        for _, subreddit := range state.Subreddits {
            if subreddit.Name == msg.Name {
                context.Respond(&messages.OperationResponse{
                    Success: false,
                    Error: "subreddit already exists",
                })
                return
            }
        }

        subredditID := fmt.Sprintf("subreddit_%d", len(state.Subreddits)+1)
        newSubreddit := &messages.SubReddit{
            Name:       msg.Name,
            Members:    make(map[string]bool),
            Moderators: map[string]bool{msg.UserId: true},
        }
        state.Subreddits[subredditID] = newSubreddit

        context.Respond(&messages.OperationResponse{
            Success: true,
            Id:      subredditID,
            Result: &messages.OperationResponse_Subreddit{
                Subreddit: newSubreddit,
            },
        })

    case *messages.JoinSubRedditMsg:
        if subreddit, exists := state.Subreddits[msg.Subreddit]; exists {
            subreddit.Members[msg.UserId] = true
            context.Respond(&messages.OperationResponse{Success: true})
        } else {
            context.Respond(&messages.OperationResponse{
                Success: false,
                Error: "subreddit not found",
            })
        }
    }
}