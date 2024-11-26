// internal/actors/simulator_actor.go
package actors

import (
    "fmt"
    "log"
    "math/rand"
    "time"
    "github.com/asynkron/protoactor-go/actor"
    "redditclone/internal/messages"
)

func (state *SimulatorActor) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *actor.Started:
        log.Println("SimulatorActor started")
        
    case *messages.StartSimulation:
        log.Printf("Starting simulation with %d users", state.NumUsers)
        rand.Seed(time.Now().UnixNano())
        go state.runSimulation(context)
        
    default:
        log.Printf("Received unknown message: %v", msg)
    }
}

func (state *SimulatorActor) runSimulation(context actor.Context) {
    log.Println("Beginning user registration phase...")

    // Register users with retry mechanism
    for i := 0; i < state.NumUsers; i++ {
        username := fmt.Sprintf("user_%d", i)
        registered := false
        retries := 3

        for attempt := 0; attempt < retries && !registered; attempt++ {
            log.Printf("Attempting to register user: %s (attempt %d)", username, attempt+1)
            
            future := context.RequestFuture(state.UserManager, 
                &messages.RegisterUserMsg{
                    Username: username,
                }, 
                10*time.Second) // Increased timeout

            if result, err := future.Result(); err != nil {
                log.Printf("Error registering user %s (attempt %d): %v", username, attempt+1, err)
                time.Sleep(time.Second) // Wait before retry
                continue
            } else {
                if response, ok := result.(*messages.OperationResponse); ok {
                    if response.Success {
                        log.Printf("Successfully registered user: %s with ID: %s", username, response.Id)
                        state.UserIDs = append(state.UserIDs, response.Id)
                        state.ActiveUsers[response.Id] = true
                        state.Stats.RegisteredUsers++
                        registered = true
                    } else {
                        log.Printf("Failed to register user %s: %s", username, response.Error)
                    }
                }
            }
        }

        if !registered {
            log.Printf("Failed to register user %s after %d attempts", username, retries)
        }

        // Add small delay between registrations to prevent overwhelming the system
        time.Sleep(100 * time.Millisecond)
    }

    log.Printf("User registration phase complete. Registered %d users", len(state.UserIDs))
    
    if len(state.UserIDs) == 0 {
        log.Printf("No users registered, stopping simulation")
        return
    }

    // Continue with the rest of the simulation...
    state.runMainSimulation(context)
}

func (state *SimulatorActor) runMainSimulation(context actor.Context) {
    // Create initial subreddits
    numSubreddits := state.NumUsers / 20 // 1 subreddit per 20 users
    if numSubreddits < 1 {
        log.Printf("Number of subreddits is <1")
        numSubreddits = 1
    }

    log.Printf("Creating %d subreddits...", numSubreddits)

    for i := 0; i < numSubreddits; i++ {
        subredditName := fmt.Sprintf("subreddit_%d", i)
        creatorID := state.UserIDs[rand.Intn(len(state.UserIDs))]
        
        future := context.RequestFuture(state.SubredditManager, 
            &messages.CreateSubRedditMsg{
                Name: subredditName,
                UserId: creatorID,
            }, 
            5*time.Second)

        if result, err := future.Result(); err == nil {
            if response, ok := result.(*messages.OperationResponse); ok && response.Success {
                state.Subreddits = append(state.Subreddits, response.Id)
                state.Stats.ActiveSubreddits++
            }
        }
    }

    // Main simulation loop
    ticker := time.NewTicker(100 * time.Millisecond)
    for range ticker.C {
        // Simulate user connections/disconnections
        state.simulateConnectivity()

        // Simulate user activities
        for userID, active := range state.ActiveUsers {
            if !active {
                continue
            }

            // Random actions
            switch rand.Intn(5) {
            case 0:
                state.simulateJoinSubreddit(context, userID)
            case 1:
                state.simulateCreatePost(context, userID)
            case 2:
                state.simulateCreateComment(context, userID)
            case 3:
                state.simulateVote(context, userID)
            case 4:
                state.simulateDirectMessage(context, userID)
            }
        }
    }
}

// Helper methods for simulation
func (state *SimulatorActor) simulateConnectivity() {
    for userID := range state.ActiveUsers {
        if rand.Float64() < 0.1 { // 10% chance to change connection state
            state.ActiveUsers[userID] = !state.ActiveUsers[userID]
        }
    }
}

func (state *SimulatorActor) simulateJoinSubreddit(context actor.Context, userID string) {
    if len(state.Subreddits) == 0 {
        return
    }
    
    subreddit := state.Subreddits[rand.Intn(len(state.Subreddits))]
    context.Request(state.SubredditManager, &messages.JoinSubRedditMsg{
        Subreddit: subreddit,
        UserId:    userID,
    })
}

func (state *SimulatorActor) simulateCreatePost(context actor.Context, userID string) {
    if len(state.Subreddits) == 0 {
        return
    }

    subreddit := state.Subreddits[rand.Intn(len(state.Subreddits))]
    future := context.RequestFuture(state.PostManager, &messages.CreatePostMsg{
        Title:     fmt.Sprintf("Post by %s", userID),
        Content:   fmt.Sprintf("Content %d", rand.Int()),
        Subreddit: subreddit,
        AuthorId:  userID,
    }, 5*time.Second)

    if result, err := future.Result(); err == nil {
        if response, ok := result.(*messages.OperationResponse); ok && response.Success {
            state.PostIDs = append(state.PostIDs, response.Id)
            state.Stats.TotalPosts++
        }
    }
}

func (state *SimulatorActor) simulateCreateComment(context actor.Context, userID string) {
    if len(state.PostIDs) == 0 {
        return
    }

    postID := state.PostIDs[rand.Intn(len(state.PostIDs))]
    parentID := ""
    if len(state.CommentIDs) > 0 && rand.Float64() < 0.3 { // 30% chance to reply to a comment
        parentID = state.CommentIDs[rand.Intn(len(state.CommentIDs))]
    }

    future := context.RequestFuture(state.CommentManager, &messages.CreateCommentMsg{
        Content:   fmt.Sprintf("Comment %d", rand.Int()),
        PostId:    postID,
        ParentId:  parentID,
        AuthorId:  userID,
    }, 5*time.Second)

    if result, err := future.Result(); err == nil {
        if response, ok := result.(*messages.OperationResponse); ok && response.Success {
            state.CommentIDs = append(state.CommentIDs, response.Id)
            state.Stats.TotalComments++
        }
    }
}

func (state *SimulatorActor) simulateVote(context actor.Context, userID string) {
    if len(state.PostIDs) == 0 && len(state.CommentIDs) == 0 {
        return
    }

    isPostVote := len(state.PostIDs) > 0 && (len(state.CommentIDs) == 0 || rand.Float64() < 0.7)
    var itemID string
    var targetActor *actor.PID

    if isPostVote {
        itemID = state.PostIDs[rand.Intn(len(state.PostIDs))]
        targetActor = state.PostManager
    } else {
        itemID = state.CommentIDs[rand.Intn(len(state.CommentIDs))]
        targetActor = state.CommentManager
    }

    context.Request(targetActor, &messages.VoteMsg{
        ItemId:   itemID,
        UserId:   userID,
        IsUpvote: rand.Float64() < 0.7, // 70% chance to upvote
    })
}

func (state *SimulatorActor) simulateDirectMessage(context actor.Context, userID string) {
    if len(state.UserIDs) < 2 {
        return
    }

    var toUserID string
    for {
        toUserID = state.UserIDs[rand.Intn(len(state.UserIDs))]
        if toUserID != userID {
            break
        }
    }

    context.Request(state.MessageManager, &messages.SendDirectMessageMsg{
        FromUserId: userID,
        ToUserId:   toUserID,
        Content:    fmt.Sprintf("Message %d", rand.Int()),
    })
    state.Stats.TotalMessages++
}