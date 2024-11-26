// cmd/simulator/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "time"
    "github.com/asynkron/protoactor-go/actor"
    "github.com/asynkron/protoactor-go/remote"
    "redditclone/internal/actors"
    "redditclone/internal/messages"
)

func main() {
    enginePort := flag.Int("engine-port", 8090, "port of the reddit engine")
    simulatorPort := flag.Int("port", 8091, "port for the simulator")
    numUsers := flag.Int("users", 10, "number of users to simulate")
    duration := flag.Duration("duration", 1*time.Minute, "duration to run the simulation")
    flag.Parse()

    system := actor.NewActorSystem()
    config := remote.Configure("127.0.0.1", *simulatorPort)
    remoting := remote.NewRemote(system, config)
    remoting.Start()

    engineAddress := fmt.Sprintf("127.0.0.1:%d", *enginePort)
    userManagerPID := actor.NewPID(engineAddress, "user-manager")

    // Create context with timeout
    context.Background()
    
    // Instead of using RequestFuture directly, create a proper sender actor
    senderProps := actor.PropsFromFunc(func(context actor.Context) {
        switch context.Message().(type) {
        case *messages.OperationResponse:
            // Handle response if needed
        }
    })
    sender := system.Root.Spawn(senderProps)

    // Send a message using the sender actor
    system.Root.Send(userManagerPID, &messages.RegisterUserMsg{
        Username: "test_user",
    })

    managerPIDs := map[string]*actor.PID{
		"user-manager":      userManagerPID,
		"subreddit-manager": actor.NewPID(engineAddress, "subreddit-manager"),
		"post-manager":      actor.NewPID(engineAddress, "post-manager"),
		"comment-manager":   actor.NewPID(engineAddress, "comment-manager"),
		"message-manager":   actor.NewPID(engineAddress, "message-manager"),
	}

    props := actor.PropsFromProducer(func() actor.Actor {
        return actors.NewSimulatorActor(*numUsers, managerPIDs)
    })

    simulator := system.Root.Spawn(props)
    system.Root.Send(simulator, &messages.StartSimulation{NumUsers: int32(*numUsers)})
    time.Sleep(*duration)

    system.Root.Stop(sender)
}