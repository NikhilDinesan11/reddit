// internal/actors/comment_manager.go
package actors

import (
    "fmt"
    //"time"
    "github.com/asynkron/protoactor-go/actor"
    "redditclone/internal/messages"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func (state *CommentManagerActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *actor.Started:
        state.UserManager = context.Parent()

    case *messages.CreateCommentMsg:
        commentID := fmt.Sprintf("comment_%d", len(state.Comments)+1)
        newComment := &messages.Comment{
            Id:        commentID,
            Content:   msg.Content,
            AuthorId:  msg.AuthorId,
            ParentId:  msg.ParentId,
            Upvotes:   0,
            Downvotes: 0,
            Timestamp: timestamppb.Now(),
        }
        
        state.Comments[commentID] = newComment

        context.Respond(&messages.OperationResponse{
            Success: true,
            Id:      commentID,
            Result: &messages.OperationResponse_Comment{
                Comment: newComment,
            },
        })

    case *messages.VoteMsg:
        if comment, exists := state.Comments[msg.ItemId]; exists {
            if msg.IsUpvote {
                comment.Upvotes++
            } else {
                comment.Downvotes++
            }
            
            // Update author's karma
            context.Request(state.UserManager, &messages.VoteMsg{
                UserId: comment.AuthorId,
                IsUpvote: msg.IsUpvote,
            })

            context.Respond(&messages.OperationResponse{Success: true})
        } else {
            context.Respond(&messages.OperationResponse{
                Success: false,
                Error: "comment not found",
            })
        }
    }
}