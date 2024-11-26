// cmd/engine/main.go
package main

import (
    "flag"
    "log"
    
    "github.com/asynkron/protoactor-go/actor"
    "github.com/asynkron/protoactor-go/remote"
    "redditclone/internal/actors"
    //"redditclone/internal/messages"
)

func main() {
    port := flag.Int("port", 8090, "port for the engine to listen on")
    flag.Parse()

    // Create the actor system
    system := actor.NewActorSystem()

    // Configure remote
    remoteConfig := remote.Configure("127.0.0.1", *port)
    remoting := remote.NewRemote(system, remoteConfig)

    // Create props for each actor type
    userManagerProps := actor.PropsFromProducer(func() actor.Actor {
        return actors.NewUserManagerActor()
    })

    subredditManagerProps := actor.PropsFromProducer(func() actor.Actor {
        return actors.NewSubRedditManagerActor()
    })

    postManagerProps := actor.PropsFromProducer(func() actor.Actor {
        return actors.NewPostManagerActor()
    })

    commentManagerProps := actor.PropsFromProducer(func() actor.Actor {
        return actors.NewCommentManagerActor()
    })

    messageManagerProps := actor.PropsFromProducer(func() actor.Actor {
        return actors.NewMessageManagerActor()
    })

    // Register actors with the remote system
    remoting.Register("user-manager", userManagerProps)
    remoting.Register("subreddit-manager", subredditManagerProps)
    remoting.Register("post-manager", postManagerProps)
    remoting.Register("comment-manager", commentManagerProps)
    remoting.Register("message-manager", messageManagerProps)

    // Start the remote system
    remoting.Start()
    log.Printf("Remote system started on port %d", *port)


    // Spawn the manager actors
    userManager := system.Root.Spawn(userManagerProps)
    subRedditManager := system.Root.Spawn(subredditManagerProps)
    postManager := system.Root.Spawn(postManagerProps)
    commentManager := system.Root.Spawn(commentManagerProps)
    messageManager := system.Root.Spawn(messageManagerProps)

    log.Printf("Reddit engine started on port %d", *port)
    log.Printf("UserManager started with PID: %v", userManager)
    log.Printf("SubReddit Manager PID: %v", subRedditManager)
    log.Printf("Post Manager PID: %v", postManager)
    log.Printf("Comment Manager PID: %v", commentManager)
    log.Printf("Message Manager PID: %v", messageManager)

    // Keep the engine running
    select {}
}