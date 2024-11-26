// internal/actors/post_manager.go
package actors

import (
    "fmt"
    //"time"
    "github.com/asynkron/protoactor-go/actor"
    "redditclone/internal/messages"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func (state *PostManagerActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *actor.Started:
        // Store the user manager PID for karma updates
        state.UserManager = context.Parent()

    case *messages.CreatePostMsg:
        postID := fmt.Sprintf("post_%d", len(state.Posts)+1)
        newPost := &messages.Post{
            Id:        postID,
            Title:     msg.Title,
            Content:   msg.Content,
            AuthorId:  msg.AuthorId,
            Subreddit: msg.Subreddit,
            Upvotes:   0,
            Downvotes: 0,
            Timestamp: timestamppb.Now(),
        }
        state.Posts[postID] = newPost

        context.Respond(&messages.OperationResponse{
            Success: true,
            Id:      postID,
            Result: &messages.OperationResponse_Post{
                Post: newPost,
            },
        })

    case *messages.VoteMsg:
        if post, exists := state.Posts[msg.ItemId]; exists {
            if msg.IsUpvote {
                post.Upvotes++
            } else {
                post.Downvotes++
            }
            
            // Update author's karma
            context.Request(state.UserManager, &messages.VoteMsg{
                UserId: post.AuthorId,
                IsUpvote: msg.IsUpvote,
            })

            context.Respond(&messages.OperationResponse{Success: true})
        } else {
            context.Respond(&messages.OperationResponse{
                Success: false,
                Error: "post not found",
            })
        }
    }
}