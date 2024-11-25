package main

import (
	"fmt"
	"sort"
	"time"
	"github.com/asynkron/protoactor-go/actor"
)

// Data structures
type User struct {
	ID       string
	Username string
	Karma    int
}

type Subreddit struct {
	Name        string
	Description string
	Members     map[string]*User
	Posts       map[string]*Post
}

type Post struct {
	ID            string
	AuthorID      string
	Title         string
	Content       string
	SubredditName string
	Comments      map[string]*Comment
	Upvotes       map[string]bool
	Downvotes     map[string]bool
	CreatedAt     time.Time
}

type Comment struct {
	ID        string
	AuthorID  string
	Content   string
	ParentID  string
	Comments  map[string]*Comment
	Upvotes   map[string]bool
	Downvotes map[string]bool
	CreatedAt time.Time
}

type DirectMessage struct {
	ID        string
	FromID    string
	ToID      string
	Content   string
	CreatedAt time.Time
}

// Actor Messages
type (
	RegisterUserMsg struct {
		Username string
	}

	CreateSubredditMsg struct {
		Name        string
		Description string
	}

	JoinSubredditMsg struct {
		UserID        string
		SubredditName string
	}

	LeaveSubredditMsg struct {
		UserID        string
		SubredditName string
	}

	CreatePostMsg struct {
		UserID        string
		SubredditName string
		Title         string
		Content       string
	}

	CreateCommentMsg struct {
		UserID   string
		ParentID string
		Content  string
	}

	VoteMsg struct {
		UserID   string
		TargetID string
		IsUpvote bool
	}

	SendDirectMessageMsg struct {
		FromID  string
		ToID    string
		Content string
	}

	GetFeedMsg struct {
		UserID string
	}

	GetDirectMessagesMsg struct {
		UserID string
	}
)

// Actor States
type RedditActor struct {
	users      map[string]*User
	subreddits map[string]*Subreddit
	directMsgs map[string][]*DirectMessage
}

// Initialize Reddit Actor
func NewRedditActor() actor.Actor {
	return &RedditActor{
		users:      make(map[string]*User),
		subreddits: make(map[string]*Subreddit),
		directMsgs: make(map[string][]*DirectMessage),
	}
}

func (state *RedditActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *RegisterUserMsg:
		userID := fmt.Sprintf("user_%d", len(state.users)+1)
		user := &User{
			ID:       userID,
			Username: msg.Username,
			Karma:    0,
		}
		state.users[userID] = user
		context.Respond(userID)

	case *CreateSubredditMsg:
		if _, exists := state.subreddits[msg.Name]; exists {
			context.Respond(fmt.Errorf("subreddit already exists"))
			return
		}

		subreddit := &Subreddit{
			Name:        msg.Name,
			Description: msg.Description,
			Members:     make(map[string]*User),
			Posts:       make(map[string]*Post),
		}
		state.subreddits[msg.Name] = subreddit
		context.Respond(nil)

	case *JoinSubredditMsg:
		subreddit, exists := state.subreddits[msg.SubredditName]
		if !exists {
			context.Respond(fmt.Errorf("subreddit not found"))
			return
		}

		user, exists := state.users[msg.UserID]
		if !exists {
			context.Respond(fmt.Errorf("user not found"))
			return
		}

		subreddit.Members[msg.UserID] = user
		context.Respond(nil)

	case *LeaveSubredditMsg:
		subreddit, exists := state.subreddits[msg.SubredditName]
		if !exists {
			context.Respond(fmt.Errorf("subreddit not found"))
			return
		}

		delete(subreddit.Members, msg.UserID)
		context.Respond(nil)

	case *CreatePostMsg:
		subreddit, exists := state.subreddits[msg.SubredditName]
		if !exists {
			context.Respond(fmt.Errorf("subreddit not found"))
			return
		}

		postID := fmt.Sprintf("post_%d", len(subreddit.Posts)+1)
		post := &Post{
			ID:            postID,
			AuthorID:      msg.UserID,
			Title:         msg.Title,
			Content:       msg.Content,
			SubredditName: msg.SubredditName,
			Comments:      make(map[string]*Comment),
			Upvotes:       make(map[string]bool),
			Downvotes:     make(map[string]bool),
			CreatedAt:     time.Now(),
		}
		subreddit.Posts[postID] = post
		context.Respond(postID)

	case *CreateCommentMsg:
		commentID := fmt.Sprintf("comment_%d", time.Now().UnixNano())
		comment := &Comment{
			ID:        commentID,
			AuthorID:  msg.UserID,
			Content:   msg.Content,
			ParentID:  msg.ParentID,
			Comments:  make(map[string]*Comment),
			Upvotes:   make(map[string]bool),
			Downvotes: make(map[string]bool),
			CreatedAt: time.Now(),
		}

		found := false
		for _, subreddit := range state.subreddits {
			// Check if parent is a post
			if post, exists := subreddit.Posts[msg.ParentID]; exists {
				post.Comments[commentID] = comment
				found = true
				break
			}

            // Check if parent is a comment using recursion
            for _, post := range subreddit.Posts {
                if state.addCommentToParent(post.Comments, msg.ParentID, comment) {
                    found = true
                    break
                }
            }
            if found {
                break
            }
		}

		if !found {
			context.Respond(fmt.Errorf("parent post or comment not found"))
			return
		}
		context.Respond(commentID)

	case *VoteMsg:
		found := false
		// Try to find and update the target (post or comment)
		for _, subreddit := range state.subreddits {
			if post, exists := subreddit.Posts[msg.TargetID]; exists {
				state.updateVotes(post.Upvotes, post.Downvotes, msg.UserID, msg.IsUpvote)
				state.updateUserKarma(post.AuthorID, msg.IsUpvote)
				found = true
				break
			}

			// Search in comments
			for _, post := range subreddit.Posts {
				if state.findAndUpdateVote(post.Comments, msg.TargetID, msg.UserID, msg.IsUpvote) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			context.Respond(fmt.Errorf("target not found"))
			return
		}
		context.Respond(nil)

	case *SendDirectMessageMsg:
		dm := &DirectMessage{
			ID:        fmt.Sprintf("dm_%d", time.Now().UnixNano()),
			FromID:    msg.FromID,
			ToID:      msg.ToID,
			Content:   msg.Content,
			CreatedAt: time.Now(),
		}

		state.directMsgs[msg.FromID] = append(state.directMsgs[msg.FromID], dm)
		state.directMsgs[msg.ToID] = append(state.directMsgs[msg.ToID], dm)
		context.Respond(dm.ID)

	case *GetFeedMsg:
		var posts []*Post
		for _, subreddit := range state.subreddits {
			if _, isMember := subreddit.Members[msg.UserID]; isMember {
				for _, post := range subreddit.Posts {
					posts = append(posts, post)
				}
			}
		}

		sort.Slice(posts, func(i, j int) bool {
			scoreI := len(posts[i].Upvotes) - len(posts[i].Downvotes)
			scoreJ := len(posts[j].Upvotes) - len(posts[j].Downvotes)
			if scoreI != scoreJ {
				return scoreI > scoreJ
			}
			return posts[i].CreatedAt.After(posts[j].CreatedAt)
		})

		context.Respond(posts)

	case *GetDirectMessagesMsg:
		messages := state.directMsgs[msg.UserID]
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].CreatedAt.After(messages[j].CreatedAt)
		})
		context.Respond(messages)
	}
}

// Helper functions
func (state *RedditActor) addCommentToParent(comments map[string]*Comment, parentID string, newComment *Comment) bool {
	if comment, exists := comments[parentID]; exists {
		comment.Comments[newComment.ID] = newComment
		return true
	}

	for _, comment := range comments {
		if state.addCommentToParent(comment.Comments, parentID, newComment) {
			return true
		}
	}
	return false
}

func (state *RedditActor) updateVotes(upvotes, downvotes map[string]bool, userID string, isUpvote bool) {
	if isUpvote {
		delete(downvotes, userID)
		upvotes[userID] = true
	} else {
		delete(upvotes, userID)
		downvotes[userID] = true
	}
}

func (state *RedditActor) updateUserKarma(authorID string, isUpvote bool) {
	if author, exists := state.users[authorID]; exists {
		if isUpvote {
			author.Karma++
		} else {
			author.Karma--
		}
	}
}

func (state *RedditActor) findAndUpdateVote(comments map[string]*Comment, targetID, userID string, isUpvote bool) bool {
	if comment, exists := comments[targetID]; exists {
		state.updateVotes(comment.Upvotes, comment.Downvotes, userID, isUpvote)
		state.updateUserKarma(comment.AuthorID, isUpvote)
		return true
	}

	for _, comment := range comments {
		if state.findAndUpdateVote(comment.Comments, targetID, userID, isUpvote) {
			return true
		}
	}
	return false
}

func main() {
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(NewRedditActor)
	pid := system.Root.Spawn(props)

	// Example usage
	context := system.Root
	
	// Register a user
	result, _ := context.RequestFuture(pid, &RegisterUserMsg{Username: "testuser"}, 5*time.Second).Result()
	userID := result.(string)
	
	// Create a subreddit
	context.RequestFuture(pid, &CreateSubredditMsg{
		Name:        "programming",
		Description: "Discussion about programming",
	}, 5*time.Second).Wait()
	
	// Join subreddit
	context.RequestFuture(pid, &JoinSubredditMsg{
		UserID:        userID,
		SubredditName: "programming",
	}, 5*time.Second).Wait()
	
	// Create a post
	postResult, _ := context.RequestFuture(pid, &CreatePostMsg{
		UserID:        userID,
		SubredditName: "programming",
		Title:         "Hello Proto.Actor",
		Content:       "This is my first post using Proto.Actor",
	}, 5*time.Second).Result()
	
	fmt.Printf("Created post with ID: %s\n", postResult)
}