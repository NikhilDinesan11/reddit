// internal/actors/types.go
package actors

import (
    "github.com/asynkron/protoactor-go/actor"
    "redditclone/internal/messages"
)

// Actor type definitions
type UserManagerActor struct {
    Users map[string]*messages.User
}

type SubRedditManagerActor struct {
    Subreddits map[string]*messages.SubReddit
}

type PostManagerActor struct {
    Posts map[string]*messages.Post
    UserManager *actor.PID
}

type CommentManagerActor struct {
    Comments map[string]*messages.Comment
    UserManager *actor.PID
}

type MessageManagerActor struct {
    Messages map[string][]*messages.DirectMessage
}

type SimulatorActor struct {
    UserIDs         []string
    Subreddits      []string
    PostIDs         []string
    CommentIDs      []string
    ActiveUsers     map[string]bool
    UserManager     *actor.PID
    SubredditManager *actor.PID
    PostManager     *actor.PID
    CommentManager  *actor.PID
    MessageManager  *actor.PID
    NumUsers        int
    Stats           *messages.SimulationStats
}

// Actor constructors
func NewUserManagerActor() *UserManagerActor {
    return &UserManagerActor{
        Users: make(map[string]*messages.User),
    }
}

func NewSubRedditManagerActor() *SubRedditManagerActor {
    return &SubRedditManagerActor{
        Subreddits: make(map[string]*messages.SubReddit),
    }
}

func NewPostManagerActor() *PostManagerActor {
    return &PostManagerActor{
        Posts: make(map[string]*messages.Post),
    }
}

func NewCommentManagerActor() *CommentManagerActor {
    return &CommentManagerActor{
        Comments: make(map[string]*messages.Comment),
    }
}

func NewMessageManagerActor() *MessageManagerActor {
    return &MessageManagerActor{
        Messages: make(map[string][]*messages.DirectMessage),
    }
}

func NewSimulatorActor(numUsers int, managerPIDs map[string]*actor.PID) *SimulatorActor {
    return &SimulatorActor{
        UserIDs:         make([]string, 0),
        Subreddits:      make([]string, 0),
        PostIDs:         make([]string, 0),
        CommentIDs:      make([]string, 0),
        ActiveUsers:     make(map[string]bool),
        NumUsers:        numUsers,
        UserManager:     managerPIDs["user-manager"],
        SubredditManager: managerPIDs["subreddit-manager"],
        PostManager:     managerPIDs["post-manager"],
        CommentManager:  managerPIDs["comment-manager"],
        MessageManager:  managerPIDs["message-manager"],
        Stats:          &messages.SimulationStats{},
    }
}